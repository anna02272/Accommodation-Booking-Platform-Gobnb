import { ComponentFixture, TestBed } from '@angular/core/testing';

import { RateAccommodationComponent } from './rate-accommodation.component';

describe('RateAccommodationComponent', () => {
  let component: RateAccommodationComponent;
  let fixture: ComponentFixture<RateAccommodationComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [ RateAccommodationComponent ]
    })
    .compileComponents();

    fixture = TestBed.createComponent(RateAccommodationComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
