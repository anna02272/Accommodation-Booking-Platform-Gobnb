import { ComponentFixture, TestBed } from '@angular/core/testing';

import { CreateAccommodationComponent } from './create-accommodation.component';

describe('CreateAccommodationComponent', () => {
  let component: CreateAccommodationComponent;
  let fixture: ComponentFixture<CreateAccommodationComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [ CreateAccommodationComponent ]
    })
    .compileComponents();

    fixture = TestBed.createComponent(CreateAccommodationComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
