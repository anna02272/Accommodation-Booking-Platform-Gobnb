import { ComponentFixture, TestBed } from '@angular/core/testing';

import { RateHostComponent } from './rate-host.component';

describe('RateHostComponent', () => {
  let component: RateHostComponent;
  let fixture: ComponentFixture<RateHostComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [ RateHostComponent ]
    })
    .compileComponents();

    fixture = TestBed.createComponent(RateHostComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
